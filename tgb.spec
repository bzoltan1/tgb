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
%gobuild ...


%install
%goinstall
install -D -m 644 tgb.yaml %{buildroot}%{_sysconfdir}/tgb.yaml
install -D -m 0644 tgb.service %{buildroot}%{_unitdir}/tgb.service

%pre
%service_add_pre tgb.service

%preun
%service_del_preun tgb.service

%post
%service_add_post tgb.service

%postun
%service_del_postun tgb.service

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_unitdir}/tgb.service
%{_bindir}/tgb
%config(noreplace) %{_sysconfdir}/tgb.yaml



%changelog
* Sat Apr 02 2021 Zoltán Balogh <zoltan@bakter.hu> - 0.1-1
- package the tgb
