package BrokerResolver

func (resolver *Resolver) GetSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Load()
}
func (resolver *Resolver) RetrieveSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Swap(0)
}

func (resolver *Resolver) GetSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Load()
}
func (resolver *Resolver) RetrieveSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Swap(0)
}

func (resolver *Resolver) GetFailedResolutions() uint64 {
	return resolver.failedResolutions.Load()
}
func (resolver *Resolver) RetrieveFailedResolutions() uint64 {
	return resolver.failedResolutions.Swap(0)
}
