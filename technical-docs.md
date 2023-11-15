# Geteduroam technical docs
Welcome to the Geteduroam technical documentation. This documentation is intended for client developers that are building a Geteduroam client. We try to cover most topics that are needed to build such a client.

This document is based on how existing clients work, including some assumptions on how they should work. Due to this, some documentation can be incorrect or incomplete. This documentation therefore still needs collaborative work, pull requests are always welcome. We want to make this document a standard to be used by Geteduroam apps, which needs agreeing on the aspects that we mention here.

Note that this documentation only describes the protocol to get the EAP metadata. Anything to do with mobile is not yet documented.

## Overview
Configuring an eduroam connection needs to happen by parsing the Extensible Authentication Protocol (EAP) metadata and then importing the metadata using the correct user provided information into the OS specific network manager.

As an overview on how a Geteduroam app/client does this is the following:

  * The app starts up
  * A file is obtained that lists each instance and their way to get the EAP metadata. This file is gotten from a discovery server, it is possible that clients implement caching using a local copy
  * This discovery file is parsed to list all the instances in it. This discovery file is parsed using JSON
  * The user selects an instance either by clicking on one or filtering on it using a search box
  * The selected instance is used to obtain the EAP metadata, either through OAuth or directly getting the configuration. It is possible that instead of getting the EAP metadata, the user is redirected to a webpage that handles the further setup
  * The EAP metadata is parsed, using XML. Validating that the EAP metadata is correct can be done through XML schemas
  * The app determines whether or not user provided credentials or secrets still need to be provided
  * When the user has entered these credentials, the eduroam profile can be configured in the OS network manager

## Discovery format
Geteduroam uses [https://discovery.eduroam.app/v1/discovery.json](https://discovery.eduroam.app/v1/discovery.json), a JSON file, to list all the institutes generated from [CAT](https://cat.eduroam.org/). This list is shown in a Geteduroam client so that the user can choose its own institution to connect to.

The script that generates this discovery JSON file can be found on the [Geteduroam GitHub](https://github.com/geteduroam/cattenbak). The format of this JSON file is the following:

```json
{
"instances": instances (list, required),
"seq": [YEAR][MONTH][DAY][UPDATE NUMBER] (integer, required); e.g. 2023020908, meaning: 2023, february 9, update 8),
"version": 1 (integer, required); always set to 1 right now,
}
```

The last part of the "seq" field is used to indicate the nth update of the day, starting at 0. If an old update cannot be found it starts at 80, see [https://github.com/geteduroam/cattenbak/blob/481e243f22b40e1d8d48ecac2b85705b8cb48494/cattenbak.py#L115](https://github.com/geteduroam/cattenbak/blob/481e243f22b40e1d8d48ecac2b85705b8cb48494/cattenbak.py#L115) for how it works in detail. It is currently undecided if clients should parse this to exactly match the format. In this client, we currently do this; meaning we parse the year month day and check for the update number. An alternative way is to parse this just as an integer without any meaning except that a newer version has a higher number.

To protect against rollback attacks, it is good practice to check the sequence number if it has updated.

Where instances is a list of the form:

```json
    "cat_idp": cat identifier (integer, required); e.g. 7088,
    "country": country code (string, required); e.g. "RO",
    "geo": [
        "lat": latitude (float, required),
        "lon": longitude (float, required),
    ] (required),
    id: cat_id (string, required); e.g. "cat_7088",
    name: the name of the organisation to be shown in the UI (string, required); e.g. SURF,
    profiles: [
        "authorization_endpoint": The authorization endpoint in case OAuth is used (string, optional); e.g. "https://example.com/oauth/authorize/",
        "default": If this profile is the default profile (bool, optional); e.g. True,
        "eapconfig_endpoint": The endpoint to obtain the EAP config (string, required); e.g. "https://example.com/api/eap-config/",
        "id": The identifier of the profile (string, required); e.g. "letswifi_cat_1337",
        "name": The name of the profile to be shown in the UI; e.g. "Demo Server",
        "oauth": Whether or not OAuth is enabled. If missing, OAuth is not enabled (bool, optional); e.g. true,
        "token_endpoint": The endpoint to get OAuth tokens from (string, optional); e.g. "https://example.com/oauth/token/", 
    ] (required),
```

This instances list should be parsed by the client. The name of the instance is what is shown in the UI. Filtering on the instance is also done with this name. For example if an user searches for "sur", it would include "SURF" due to substring case-insensitive matching.

### Variants/flows
As can be deduced from the instance format, there are various flows possible to configure the eduroam network with a certain profile:

  * Directly get the EAP config from the `eapconfig_endpoint`: `eapconfig_endpoint`, `name`, `id` and `oauth` (oauth set to False) MUST be present
  * Get the EAP config using tokens obtained through `authorization_endpoint` and `token_endpoint` using OAuth: `eapconfig_endpoint`, `authorization_endpoint`, `token_endpoint`, `name`, `id` and `oauth` (oauth set to True) MUST be present
  * Forward the user to a "redirect" page: `id`, `name` and `redirect` must be present

Based on the various presence and values of these attributes you can determine the flow as follows:
  * If `redirect` is present, then the redirect flow MUST be used
  * Else, check whether `oauth` is set to True then OAuth flow, else direct flow
    - For a more complete check, instead of only checking if `oauth` is True a client can also check for the presence of `authorization_endpoint` and `token_endpoint`

The implementation of each flow will be given later. Before we can do that, however, we first explain how a profile should be selected

## Profile selection
As can be deduced from the JSON format, there are multiple profiles available per instance. If an instance only has one profile then the profile MUST be automatically chosen without any user interaction. 

If there are multiple profiles then multiple profiles MUST be shown in the UI, asking for a selection to the user. The profile indicated with the `default` attribute set to true SHOULD be in bold, or in case the UI does not support bold text, it SHOULD have a */(default) pre/postfix.

When the profile has been selected, we can use the correct flow to get the EAP metadata. In the next section, we will go over implementing the various flows.

## Flow implementations
This section describes the different way that the app should continue when the profile has been selected.

### Redirect
After parsing the discovery entry and determining that the flow is Redirect, the redirect should be verified whether or not the following holds:

  * The value is a URL
  * The scheme of the URL is HTTPS or HTTP

If the value is not a URL, or the scheme is not HTTP/HTTPS, the app MUST NOT open the url in the browser but should show a friendly error in the UI that the profile is not available.

If the scheme of the URL is HTTP it MUST be rewritten to HTTPS.

Note that the redirect flow is one of the last steps that the app needs to do as the redirect does not give back an EAP metadata file. This redirect is only used to give the user information on how to proceed with configuring the network himself.

### Direct
When the app has determined that the profile does not support redirect and oauth is disabled, the app should get the eap config via the `eapconfig_endpoint`. The EAP metadata file is returned in the HTTP response body. 

Note that like the URL in redirect, the app MUST parse the `eapconfig_endpont` to check whether or not it is a valid URL, the scheme is HTTP or HTTPS and MUST rewrite HTTP to the HTTPS scheme.

### OAuth
The extra fields that are available in the OAuth flow are the `authorization_endpoint` and the `token_endpoint`. We go over them one by one what should be done.

NOTE: The authorization endpoint and token endpoint docs is taken from https://www.geteduroam.app/developer/api/ and slightly modified

#### Authorization endpoint

Build a URL for the authorization endpoint; take the `authorization_endpoint` string from the discovery,
and add the following GET parameters (MUST be implemented according to RFC6749 section 4.1.1 for most of these):

  * `response_type` (MUST be set to `code`)
  * `code_challenge_method` (MUST be set to `S256`)
  * `scope` (MUST be set to `eap-metadata`)
  * `code_challenge` (a code challenge)
  * `redirect_uri` (where the user should be redirected after accepting or rejecting your application, GET parameters will added to this URL by the server. MUST be local, e.g. http://127.0.0.1/callback)
  * `client_id` (MUST be your client ID as known by the server)
  * `state` (a random string that will be set in a GET parameter to the redirect_uri, for you to verify it’s the same flow))

You have created a URL, for example:

	https://demo.eduroam.no/authorize.php?response_type=code&code_challenge_method=S256&scope=eap-metadata&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&redirect_uri=http%3A%2F%2Flocalhost%3A1080%2Fcallback&client_id=00000000-0000-0000-0000-000000000000&state=0

You open a local webbrowser to this URL on the users' device and listen on the `redirect_uri` for a request to return.
Upon receiving a request, the client SHOULD reclaim focus to the application window and MUST handle the request.
You may receive these GET parameters:

  * `code` (the authorization code that you can use on the token endpoint)
  * `error` (an error message that you can present to the user)
  * `state` (the same value as your earlier `state` GET parameter which MUST be checked)

As a reply to this request, you SHOULD simply return a message to the user stating that he should return to the application.
Depending on the platform, you SHOULD also return code to trigger a return to the application.

#### Token endpoint

The token endpoint requires a `code`, which you obtain via the Authorization endpoint.
Use the `token_endpoint` string from the discovery.

You need the following POST parameters:

  * `grant_type` (MUST be set to `authorization_code`)
  * `code` (MUST be the code received from the authorization endpoint)
  * `redirect_uri` (MUST repeat the value used in the previous request, as mandated by RFC7636)
  * `client_id` (MUST repeat the value used in the previous request, as mandated by RFC7636)
  * `code_verifier` (MUST be a code verifier, as documented in RFC7636 section 4. This is the preimage of the code challenge to prove that you are the original sender of the authorization endpoint request )

You get back a JSON dictionary, containing the following keys:

  * `access_token`
  * `token_type` (set to `Bearer`)
  * `expires_in` (validity of the `access_token` in seconds)

Example HTTP conversation

	POST /token.php HTTP/1.1
	Accept: application/json
	Content-Type: application/x-www-form-urlencoded
	Content-Length: 209

	grant_type=authorization_code&code=v2.local.AAAAAA&redirect_uri=http%3A%2F%2Flocalhost%3A1080%2Fcallback&client_id=00000000-0000-0000-0000-000000000000&code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk

	HTTP/1.1 200 OK
	Cache-Control: no-store
	Content-Type: application/json;charset=UTF-8
	Pragma: no-cache

	{
		"access_token": "v2.local.AAAAA…==",
		"token_type": "Bearer",
		"expires_in": 3600
	}

Saving this access token SHOULD be done securely, e.g. in a keyring. This way the client can reuse this access token across restarts.

#### Doing the authorized request
Now that the client has retrieved the access token, it needs to get the EAP metadata using it. To do this, the client MUST send the access token in the authorization header when making a request to `eapconfig_endpoint`:

	curl \
		-H "Authorization: Bearer SETTHETOKENHERE" \
		https://example.org/api/eap-config

Note that error handling on the HTTP code should done to accordance with RFC6749. In short, when the client gets a HTTP 401 here then that possibly means that the access token is expired or invalid/blacklisted. Therefore the client MUST check before it sends the request if the access token is still valid.

If the 401 is returned, or the client did not even have the access token in the first place the whole OAuth procedure MUST be redone.
