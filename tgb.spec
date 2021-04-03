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
BuildRequires:  golang(API) = 1.14
BuildRequires:  golang(github.com/gorilla/mux) >= 0-0.13
BuildRequires:  golang(github.com/go-telegram-bot-api/telegram-bot-api)
BuildRequires:  golang(gopkg.in/yaml.v2)
BuildRequires:  golang(fmt)
BuildRequires:  golang(log)
BuildRequires:  golang(encoding/json)
BuildRequires:  golang(net/http)
BuildRequires:  golang(io/ioutil)


%description
Telegram bridge service for any user or process to send warnings, alerts
to a configured Telegram user.


%prep
%setup -q -n tgb-%{version}

# many golang binaries are "vendoring" (bundling) sources, so remove them. Those dependencies need to be packaged independently.
rm -rf vendor

%build
# set up temporary build gopath, and put our directory there
mkdir -p ./_build/src/github.com/bzoltan1
ln -s $(pwd) ./_build/src/github.com/bzoltan1/tgb

export GOPATH=$(pwd)/_build:%{gopath}
go build -o tgb .

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./tgb %{buildroot}%{_bindir}/tgb

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/tgb

%changelog
* Sat Apr 02 2021 Zoltán Balogh <zoltan@bakter.hu> - 0.1-1
- package the tgb
