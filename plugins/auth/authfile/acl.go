package acl

type aclAuth struct {
	config * ACLConfig
}

func Init() *aclAuth {
	aclConfig,err := AclConfigLoad("./plugins/auth/authfile/acl.conf")
	if err != nil {
		panic(err)
	}
	return &aclAuth{
		config: aclConfig,
	}
}
