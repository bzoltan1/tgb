%go_nostrip

%define _buildshell /bin/bash
%define import_path github.com/bzoltan1/tgb

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

%description
Telegram bridge service for any user or process to send warnings, alerts
to a configured Telegram user.


%prep
%setup -q -n tgb-%{version}

# many golang binaries are "vendoring" (bundling) sources, so remove them. Those dependencies need to be packaged independently.
rm -rf vendor

%build
%goprep %{import_path}
%gobuild ...


%install
%goinstall
%gosrc
%gofilelist

%check
%gotest %{import_path}...

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/tgb

%changelog
* Sat Apr 02 2021 Zoltán Balogh <zoltan@bakter.hu> - 0.1-1
- package the tgb
