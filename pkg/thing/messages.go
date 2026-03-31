package thing

const (
	FieldCannotBeEmpty           = "field %s cannot be empty or contain only spaces"
	FieldMinLengthIsN            = "field %s minimum length is %d"
	FoundNum                     = ", found %d"
	NoRowsAffectedInFunc         = "no row where affected during %s"
	FunctionNReturnedNoResults   = "%s returned no results "
	OnlyAdminCanManageTypeThings = "only admin user can manage type thing"
	SelectFailedInNWithErrorE    = "pgxscan.Select unexpectedly failed in %s, error : %v"
)
