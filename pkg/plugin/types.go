package plugin

type (
	BloblangFunction func() (any, error)

	BloblangMapping interface {
		Query(any) (any, error)
	}
)
