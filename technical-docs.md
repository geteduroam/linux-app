# Geteduroam technical docs
Welcome to the Geteduroam technical documentation. This documentation is intended for client developers that are building a Geteduroam client. We try to cover most topics that are needed to build such a client.

This document is based on how existing clients work, including some assumptions on how they should work. Due to this, some documentation can be incorrect or incomplete. This documentation therefore still needs collaborative work, pull requests are always welcome. We want to make this document a standard to be used by Geteduroam apps, which needs agreeing on the aspects that we mention here.

## Overview
Configuring an eduroam connection needs to happen by parsing the Extensible Authentication Protocol (EAP) metadata and then importing the metadata using the correct user provided information into the OS specific network manager.

As an overview on how a Geteduroam does this is the following:

- The app starts up
- A file is obtained that lists each instance and their way to get the EAP metadata. This file is gotten from a discovery server, it is possible that clients implement caching using a local copy
- This discovery file is parsed to list all the instances in it. This discovery file is parsed using JSON
- The user selects an instance either by clicking on one or filtering on it using a search box
- The selected instance is used to obtain the EAP metadata, either through OAuth or directly getting the configuration. It is possible that instead of getting the EAP metadata, the user is redirected to a webpage that handles the further setup
- The EAP metadata is parsed, using XML. Validating that the EAP metadata is correct can be done through XML schemas
- The app determines whether or not user provided credentials or secrets still need to be provided
- When the user has entered these credentials, the eduroam profile can be configured in the OS network manager

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

- Directly get the EAP config from the `eapconfig_endpoint`: `eapconfig_endpoint`, `name`, `id` and `oauth` (oauth set to False) MUST be present
- Get the EAP config using tokens obtained through `authorization_endpoint` and `token_endpoint` using OAuth: `eapconfig_endpoint`, `authorization_endpoint`, `token_endpoint`, `name`, `id` and `oauth` (oauth set to True) MUST be present
- Forward the user to a "redirect" page: `id`, `name` and `redirect` must be present

Based on the various presence and values of these attributes you can determine the flow as follows:
- If `redirect` is present, then the redirect flow MUST be used
- Else, check whether `oauth` is set to True then OAuth flow, else direct flow
  - For a more complete check, instead of only checking if `oauth` is True a client can also check for the presence of `authorization_endpoint` and `token_endpoint`

The implementation of each flow will be given in the next section

## Flow implementations

### Redirect
After parsing the discovery entry and determining that the flow is Redirect, the redirect should be verified whether or not the following holds:

- The value is an URL
- The scheme of the URL is HTTPS

If the value is not an URL, or the scheme is HTTP (an insecure URL), the app SHOULD NOT open the url in the browser but should show a friendly error in the UI that the profile is not available.
