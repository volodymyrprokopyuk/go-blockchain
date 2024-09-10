package rtx

type TxSrv struct {
  UnimplementedTxServer
  keyStoreDir string
}

func NewTxSrv(keyStoreDir string) *TxSrv {
  return &TxSrv{keyStoreDir: keyStoreDir}
}
