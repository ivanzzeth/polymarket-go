package polymarket

// DataSource specifies the data source for queries
type DataSource int

const (
	// DataSourceCLOB queries from CLOB API (default, may include additional info like locked amounts)
	DataSourceCLOB DataSource = iota
	// DataSourceOnChain queries directly from blockchain (source of truth)
	DataSourceOnChain
)

// String returns the string representation of the data source
func (d DataSource) String() string {
	switch d {
	case DataSourceCLOB:
		return "CLOB"
	case DataSourceOnChain:
		return "OnChain"
	default:
		return "Unknown"
	}
}
