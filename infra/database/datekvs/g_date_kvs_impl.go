package datekvs

//type dateKvsImpl[T any] struct {
//	rawKvs   kvstore.RawJsonStore
//	storName string
//}
//
//func NewDateKvsImpl[T any](rawKvs kvstore.RawJsonStore, storName string) GlobalDateKvStore[T] {
//	return &dateKvsImpl[T]{rawKvs: rawKvs, storName: storName}
//}
//
//func (d *dateKvsImpl[T]) getGlobalStore(ctx context.Context, key string, forDate Date) (kvstore.GlobalKvStore[T], error) {
//	keyStore, err := d.getStoreForKey(ctx, key)
//	if err != nil {
//		return nil, err
//	}
//	return getDateStore[T](keyStore, forDate), nil
//
//}
//
//func getDateStore[T any](keyStore kvstore.RawJsonStore, forDate Date) kvstore.GlobalKvStore[T] {
//	return kvstore.NewGlobalJsonKvStoreImpl[T](keyStore, fmt.Sprintf("%d", forDate.Year()))
//}
//
//func (d *dateKvsImpl[T]) getStoreForKey(ctx context.Context, key string) (kvstore.RawJsonStore, error) {
//	keyStore, err := d.rawKvs.CreateSpaceStore(ctx, []string{d.storName, key})
//	if err != nil {
//		return nil, err
//	}
//	return keyStore, nil
//}
//func (d *dateKvsImpl[T]) Set(ctx context.Context, key string, forDate Date, value T) error {
//	g, err := d.getGlobalStore(ctx, key, forDate)
//	if err != nil {
//		return err
//	}
//	return g.Set(ctx, value)
//}
//
//func (d *dateKvsImpl[T]) Unset(ctx context.Context, key string, forDate Date) error {
//	g, err := d.getGlobalStore(ctx, key, forDate)
//	if err != nil {
//		return err
//	}
//	return g.Unset(ctx)
//}
//
//func (d *dateKvsImpl[T]) Get(ctx context.Context, key string, forDate Date) (T, error) {
//	g, err := d.getGlobalStore(ctx, key, forDate)
//	if err != nil {
//		var ret T
//		return ret, err
//	}
//	return g.Get(ctx)
//}
//
//func (d *dateKvsImpl[T]) Find(ctx context.Context, key string, forDate Date) (*T, error) {
//	g, err := d.getGlobalStore(ctx, key, forDate)
//	if err != nil {
//		return nil, err
//	}
//	return g.GetIfExist(ctx)
//}
//
//func (d *dateKvsImpl[T]) GetRange(ctx context.Context, key string, from Date, to Date) (shpanstream.Stream[DatedRecord[T]], error) {
//	forKey, err := d.getStoreForKey(ctx, key)
//	if err != nil {
//		return nil, err
//	}
//
//	// todo:amit:can use "list" to make more efficient probably instead of every day
//	currTime := from
//	return shpanstream.NewSimpleStream[DatedRecord[T]](func(ctx context.Context) (*DatedRecord[T], error) {
//		for !currTime.After(to.Time) {
//			exist, err := getDateStore[T](forKey, currTime).GetIfExist(ctx)
//			if err != nil {
//				return nil, err
//			}
//			if exist != nil {
//				return &DatedRecord[T]{currTime, *exist}, nil
//			}
//			currTime = currTime.AddDay()
//		}
//		return nil, io.EOF
//
//	}), nil
//}
