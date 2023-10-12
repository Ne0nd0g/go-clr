//go:build windows
// +build windows

package clr

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ICORRuntimeHost struct {
	vtbl *ICORRuntimeHostVtbl
}

// ICORRuntimeHostVtbl Provides methods that enable the host to start and stop the common language runtime (CLR)
// explicitly, to create and configure application domains, to access the default domain, and to enumerate all
// domains running in the process.
// https://docs.microsoft.com/en-us/dotnet/framework/unmanaged-api/hosting/icorruntimehost-interface
type ICORRuntimeHostVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	// CreateLogicalThreadState Do not use.
	CreateLogicalThreadState uintptr
	// DeleteLogicalThreadSate Do not use.
	DeleteLogicalThreadState uintptr
	// SwitchInLogicalThreadState Do not use.
	SwitchInLogicalThreadState uintptr
	// SwitchOutLogicalThreadState Do not use.
	SwitchOutLogicalThreadState uintptr
	// LocksHeldByLogicalThreadState Do not use.
	LocksHeldByLogicalThreadState uintptr
	// MapFile Maps the specified file into memory. This method is obsolete.
	MapFile uintptr
	// GetConfiguration Gets an object that allows the host to specify the callback configuration of the CLR.
	GetConfiguration uintptr
	// Start Starts the CLR.
	Start uintptr
	// Stop Stops the execution of code in the runtime for the current process.
	Stop uintptr
	// CreateDomain Creates an application domain. The caller receives an interface pointer of
	// type _AppDomain to an instance of type System.AppDomain.
	CreateDomain uintptr
	// GetDefaultDomain Gets an interface pointer of type _AppDomain that represents the default domain for the current process.
	GetDefaultDomain uintptr
	// EnumDomains Gets an enumerator for the domains in the current process.
	EnumDomains uintptr
	// NextDomain Gets an interface pointer to the next domain in the enumeration.
	NextDomain uintptr
	// CloseEnum Resets a domain enumerator back to the beginning of the domain list.
	CloseEnum uintptr
	// CreateDomainEx Creates an application domain. This method allows the caller to pass an
	// IAppDomainSetup instance to configure additional features of the returned _AppDomain instance.
	CreateDomainEx uintptr
	// CreateDomainSetup Gets an interface pointer of type IAppDomainSetup to an AppDomainSetup instance.
	// IAppDomainSetup provides methods to configure aspects of an application domain before it is created.
	CreateDomainSetup uintptr
	// CreateEvidence Gets an interface pointer of type IIdentity, which allows the host to create security
	// evidence to pass to CreateDomain or CreateDomainEx.
	CreateEvidence uintptr
	// UnloadDomain Unloads the specified application domain from the current process.
	UnloadDomain uintptr
	// CurrentDomain Gets an interface pointer of type _AppDomain that represents the domain loaded on the current thread.
	CurrentDomain uintptr
}

// GetICORRuntimeHost is a wrapper function that takes in an ICLRRuntimeInfo and returns an ICORRuntimeHost object
// and loads it into the current process. This is the "deprecated" API, but the only way currently to load an assembly
// from memory (afaict)
func GetICORRuntimeHost(runtimeInfo *ICLRRuntimeInfo) (*ICORRuntimeHost, error) {
	debugPrint("Entering into icorruntimehost.GetICORRuntimeHost()...")
	var runtimeHost *ICORRuntimeHost
	err := runtimeInfo.GetInterface(CLSID_CorRuntimeHost, IID_ICorRuntimeHost, unsafe.Pointer(&runtimeHost))
	if err != nil {
		return nil, err
	}

	err = runtimeHost.Start()
	return runtimeHost, err
}

func (obj *ICORRuntimeHost) QueryInterface(riid windows.GUID, ppvObject unsafe.Pointer) error {
	debugPrint("Entering into icorruntimehost.QueryInterface()...")
	hr, _, err := syscall.Syscall(
		obj.vtbl.QueryInterface,
		3,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(&riid)), // A reference to the interface identifier (IID) of the interface being queried for.
		uintptr(ppvObject),
	)
	if err != syscall.Errno(0) {
		fmt.Println("1111111111111")
		return fmt.Errorf("the IUknown::QueryInterface method returned an error:\r\n%s", err)
	}
	if hr != S_OK {
		fmt.Println("222222222222222222")
		return fmt.Errorf("the IUknown::QueryInterface method method returned a non-zero HRESULT: 0x%x", hr)
	}
	return nil
}

func (obj *ICORRuntimeHost) AddRef() uintptr {
	ret, _, _ := syscall.Syscall(
		obj.vtbl.AddRef,
		1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)
	return ret
}

func (obj *ICORRuntimeHost) Release() uintptr {
	ret, _, _ := syscall.Syscall(
		obj.vtbl.Release,
		1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)
	return ret
}

// Start starts the common language runtime (CLR).
// HRESULT Start ();
// https://docs.microsoft.com/en-us/dotnet/framework/unmanaged-api/hosting/icorruntimehost-start-method
func (obj *ICORRuntimeHost) Start() error {
	debugPrint("Entering into icorruntimehost.Start()...")
	hr, _, err := syscall.Syscall(
		obj.vtbl.Start,
		1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0,
	)
	if err != syscall.Errno(0) {
		// The system could not find the environment option that was entered.
		// TODO Why is this error message returned?
		debugPrint(fmt.Sprintf("the ICORRuntimeHost::Start method returned an error:\r\n%s", err))
	}
	if hr != S_OK {
		return fmt.Errorf("the ICORRuntimeHost::Start method method returned a non-zero HRESULT: 0x%x", hr)
	}
	return nil
}

// GetDefaultDomain gets an interface pointer of type System._AppDomain that represents the default domain for the current process.
// HRESULT GetDefaultDomain (
//
//	[out] IUnknown** pAppDomain
//
// );
// https://docs.microsoft.com/en-us/dotnet/framework/unmanaged-api/hosting/icorruntimehost-getdefaultdomain-method
func (obj *ICORRuntimeHost) GetDefaultDomain() (IUnknown *IUnknown, err error) {
	debugPrint("Entering into icorruntimehost.GetDefaultDomain()...")
	hr, _, err := syscall.Syscall(
		obj.vtbl.GetDefaultDomain,
		2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(&IUnknown)),
		0,
	)
	if err != syscall.Errno(0) {
		// The specified procedure could not be found.
		// TODO Why is this error message returned?
		debugPrint(fmt.Sprintf("the ICORRuntimeHost::GetDefaultDomain method returned an error:\r\n%s", err))
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::GetDefaultDomain method method returned a non-zero HRESULT: 0x%x", hr)
		return
	}
	err = nil
	return
}

// CreateDomain Creates an application domain. The caller receives an interface pointer of type _AppDomain to an instance of type System.AppDomain.
// HRESULT CreateDomain (
//
//	[in] LPWSTR    pwzFriendlyName,
//	[in] IUnknown* pIdentityArray,
//	[out] void   **pAppDomain
//
// );
// https://docs.microsoft.com/en-us/previous-versions/dotnet/netframework-4.0/ms164322(v=vs.100)
func (obj *ICORRuntimeHost) CreateDomain(FriendlyName string) (pAppDomain *AppDomain, err error) {
	pwzFriendlyName := &utf16Le(FriendlyName)[0]
	var iu *IUnknown
	debugPrint("Entering into icorruntimehost.CreateDomain()...")
	hr, _, err := syscall.Syscall6(
		obj.vtbl.CreateDomain,
		4,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pwzFriendlyName)), // [in] LPWSTR    pwzFriendlyName - An optional parameter used to give a friendly name to the domain
		uintptr(unsafe.Pointer(nil)),             // [in] IUnknown* pIdentityArray - An optional array of pointers to IIdentity instances that represent evidence mapped through security policy to establish a permission set
		uintptr(unsafe.Pointer(&iu)),             // [out] IUnknown** pAppDomain
		0,
		0,
	)
	if err != syscall.Errno(0) {
		// The specified procedure could not be found.
		// TODO Why is this error message returned?
		debugPrint(fmt.Sprintf("the ICORRuntimeHost::CreateDomain method returned an error:\r\n%s", err))
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::CreateDomain method returned a non-zero HRESULT: 0x%x", hr)
		return
	}

	err = iu.QueryInterface(IID_AppDomain, unsafe.Pointer(&pAppDomain))
	return
}

func (obj *ICORRuntimeHost) GetDomain(dName string) (pAppDomain *AppDomain, err error) {
	hEnum, err := obj.EnumDomains()
	if err != nil {
		return
	}
	for {
		ad, err := obj.NextDomain(hEnum)
		if err != nil {
			if strings.HasSuffix(err.Error(), "0x1") {
				break
			}
			return nil, err
		}
		thisName, err := ad.GetFriendlyName()
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(dName, thisName) {
			return ad, nil
		}
	}
	return nil, fmt.Errorf("could not find domain: %s", dName)
}

// EnumDomains Gets an enumerator for the domains in the current process.
// HRESULT EnumDomains (
//
//	[out] HCORENUM *hEnum
//
// );
func (obj *ICORRuntimeHost) EnumDomains() (hEnum windows.Handle, err error) {
	debugPrint("Entering into icorruntimehost.EnumDomains()...")

	hr, _, err := syscall.SyscallN(
		obj.vtbl.EnumDomains,
		0, //I don't know why
		(uintptr(unsafe.Pointer(&hEnum))),
	)

	if err != syscall.Errno(0) {
		err = fmt.Errorf("the ICORRuntimeHost::EnumDomains method returned an error:\n%s", err)
		return
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::EnumDomains method returned a non-zero HRESULT: 0x%x", hr)
		return
	}
	err = nil
	return
}

func (obj *ICORRuntimeHost) NextDomain(hDomainEnum windows.Handle) (ad *AppDomain, err error) {
	debugPrint("Entering into icorruntimehost.NextDomain()...")
	var iu *IUnknown
	hr, _, err := syscall.SyscallN(
		obj.vtbl.NextDomain,
		uintptr(unsafe.Pointer(obj)),
		uintptr(hDomainEnum),
		uintptr(unsafe.Pointer(&iu)),
	)
	if err != syscall.Errno(0) {
		err = fmt.Errorf("the ICORRuntimeHost::NextDomain method returned an error:\n%s", err)
		return
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::NextDomain method returned a non-zero HRESULT: 0x%x", hr)
		return
	}
	err = iu.QueryInterface(IID_AppDomain, unsafe.Pointer(&ad))

	return
}

func (obj *ICORRuntimeHost) CloseEnum(hDomainEnum windows.Handle) (err error) {
	debugPrint("Entering into icorruntimehost.CloseEnum()...")

	hr, _, err := syscall.SyscallN(
		obj.vtbl.CloseEnum,
		uintptr(unsafe.Pointer(obj)),
		uintptr(hDomainEnum),
	)
	if err != syscall.Errno(0) {
		err = fmt.Errorf("the ICORRuntimeHost::CloseEnum method returned an error:\n%s", err)
		return err
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::CloseEnum method returned a non-zero HRESULT: 0x%x", hr)
		return err
	}
	err = nil
	return err
}

func (obj *ICORRuntimeHost) UnloadDomain(appdomain *AppDomain) (err error) {
	debugPrint("Entering into icorruntimehost.UnloadDomain()...")

	hr, _, err := syscall.SyscallN(
		obj.vtbl.UnloadDomain,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(appdomain)),
	)
	if err != syscall.Errno(0) {
		err = fmt.Errorf("the ICORRuntimeHost::UnloadDomain method returned an error:\n%s", err)
		return err
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::UnloadDomain method returned a non-zero HRESULT: 0x%x", hr)
		return err
	}
	err = nil
	return err
}

func (obj *ICORRuntimeHost) Stop() (err error) {
	debugPrint("Entering into icorruntimehost.Stop()...")

	hr, _, err := syscall.SyscallN(
		obj.vtbl.Stop,
		uintptr(unsafe.Pointer(obj)),
	)
	if err != syscall.Errno(0) {
		err = fmt.Errorf("the ICORRuntimeHost::UnloadDomain method returned an error:\n%s", err)
		return err
	}
	if hr != S_OK {
		err = fmt.Errorf("the ICORRuntimeHost::UnloadDomain method returned a non-zero HRESULT: 0x%x", hr)
		return err
	}
	err = nil
	return err
}
