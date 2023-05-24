# https://github.com/geteduroam/linux-app

%global goipath         github.com/geteduroam/linux-app
%global commit          010aee553c263d4bac278f7a42bd850c39bfa5b3

BuildRequires: golang(github.com/godbus/dbus/v5)
BuildRequires: golang(golang.org/x/term)
BuildRequires: golang(golang.org/x/text/runes)
BuildRequires: golang(golang.org/x/text/transform)
BuildRequires: golang(golang.org/x/text/unicode/norm)
Requires: NetworkManager

Source0: https://github.com/geteduroam/linux-app/archive/%{commit}.tar.gz

%gometa -f

%global golicenses      LICENSE
%global godocs          README.md technical-docs.md

Name:           %{goname}
Version:        0.0.1
Release:        %autorelease -p
Summary:        None
License:        BSD-3-Clause
URL:            %{gourl}
Source:         %{gosource}

%description 
Geteduroam Linux CLI client

%gopkg

%prep
%goprep
%autopatch -p1
%generate_buildrequires
%go_generate_buildrequires

%build
%gobuild -o %{gobuilddir}/bin/geteduroam-cli %{goipath}/cmd/geteduroam

%install
%gopkginstall
install -m 0755 -vd                     %{buildroot}%{_bindir}
install -m 0755 -vp %{gobuilddir}/bin/* %{buildroot}%{_bindir}/

%if %{with check}
%check
%gocheck
%endif

%files
%license LICENSE
%doc README.md technical-docs.md

%{_bindir}/*

%gopkgfiles

%changelog
%autochangelog
