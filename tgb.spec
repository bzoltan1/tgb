%go_nostrip

%define _buildshell /bin/bash
%global provider        github
%global provider_tld    com
%global project         bzoltan1
%global repo            tgb
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}

Name:           tgb
Version:        0.1
Release:        0
Summary:        Telegram bridge service
License:        LGPL-2.1-or-later
Group:          System/Management
URL:            https://github.com/bzoltan1/trg
Source:         %{name}-%{version}.tar.gz
Source:         vendor.tar.gz
BuildRequires:  golang-packaging

%{go_nostrip}
%{go_provides}

%{go_exclusivearch}

%description
Telegram bridge service for any user or process to send warnings, alerts
to a configured Telegram user.

%prep
%setup -q -n %{repo}-%{version}


%build
%goprep %{import_path}
%gobuild -mod=vendor "" ...


%install
%goinstall
%gosrc
%gofilelist

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/tgb

%changelog
* Sat Apr 02 2021 Zoltán Balogh <zoltan@bakter.hu> - 0.1-1
- package the tgb
