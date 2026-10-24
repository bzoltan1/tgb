#
# spec file for package name
#
# Copyright (c) 2022 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


%define goflags "-buildmode=pie -trimpath -mod=vendor -modcacherw"
%define goipath    %{_libexecdir}/go
Name:           tgb
Version:        0.2
Release:        0
Summary:        Telegram bridge service
License:        LGPL-2.1-or-later
Group:          System/Management
URL:            https://github.com/bzoltan1/tgb
Source0:        %{name}-%{version}.tar.gz
Source1:        vendor.tar.gz
BuildRequires:  golang(API) >= 1.18
#%{go_nostrip}

%description
Telegram bridge service for any user or process to send warnings, alerts
to a configured Telegram user.

%prep
%autosetup -n tgb-%{version} -a 1

%build
go build -mod=vendor tgb

%install
install -D -m0755 %{_builddir}/%{name}-%{version}/%{name} %{buildroot}%{_bindir}/%{name}
#install -D -m755 %{name} %{buildroot}%{goipath}/%{name}
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
