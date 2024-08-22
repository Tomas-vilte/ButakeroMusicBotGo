package decoder

import "errors"

var (
	ErrNotDCA            = errors.New("No se encontró el encabezado mágico DCA, puede que no sea un archivo DCA o que contenga tramas DCA en bruto")
	ErrNotFirstFrame     = errors.New("La metadata solo puede encontrarse en el primer marco")
	ErrInvalidMetaLen    = errors.New("Longitud de metadata inválida")
	ErrNegativeFrameSize = errors.New("Tamaño del marco es negativo, posiblemente está corrupto")
)
