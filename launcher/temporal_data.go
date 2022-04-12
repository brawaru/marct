package launcher

func (w *Instance) GetTempValue(key any) any {
	return w.TemporalData[key]
}

func (w *Instance) SetTempValue(key any, v any) {
	w.TemporalData[key] = v
}
