package business

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	structs "github.com/fatih/structs"
	dbr "github.com/gocraft/dbr"
	graphql "github.com/graph-gophers/graphql-go"
	b "github.com/pickjunk/bgo"
	bt "github.com/pickjunk/bgo/time"
)

func init() {
	Graphql.MergeSchema(`
	type Query {
		admins(page: Int, search: String): AdminList
		admin(id: ID!): Admin
	}

	type Mutation {
		upsertAdmin(id: ID!, data: AdminInput!): Boolean!
		banAdmins(ids: [ID!]!, status: Boolean!): Boolean!
		deleteAdmins(ids: [ID!]!): Boolean!

		updateProfile(data: AdminInput!): Boolean!
	}

	type AdminList {
		list: [Admin!]!
		total: Int!
	}

	type Admin {
		id: ID
		name: String
		ctime: String
		mtime: String
		ltime: String
		btime: String
	}

	input AdminInput {
		name: String
		passwd: String
	}
	`)
}

// Admin struct
type Admin struct {
	ID     *graphql.ID
	Name   *string
	Passwd *string
	Ctime  dbr.NullTime
	Mtime  dbr.NullTime
	Ltime  dbr.NullTime
	Btime  dbr.NullTime
}

// AdminInput struct
type AdminInput struct {
	Name   *string
	Passwd *string
}

// AdminListResolver struct
type AdminListResolver struct {
	page   int32
	search string
	Count  int32
}

// AdminResolver struct
type AdminResolver struct {
	admin *Admin
}

// Admins resolver
func (r *resolver) Admins(
	ctx context.Context,
	args struct {
		Page   *int32
		Search *string
	},
) *AdminListResolver {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	id := ctx.Value(b.CtxKey("id")).(string)

	if id != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return nil
	}

	var lr AdminListResolver

	if args.Page == nil || *args.Page < 1 {
		lr.page = 1
	} else {
		lr.page = *args.Page
	}

	if args.Search != nil {
		lr.search = *args.Search
	}

	return &lr
}

// List resolver
func (r *AdminListResolver) List(ctx context.Context) []AdminResolver {
	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)

	var l []AdminResolver

	var size int32 = 20
	builder := db.Select("*").
		From("admin").
		Where("dtime IS NULL").
		OrderBy("mtime DESC").
		Offset(uint64((r.page - 1) * size)).
		Limit(uint64(size))

	if r.search != "" {
		builder.Where(dbr.Or(
			dbr.Expr("name LIKE ?", "%"+r.search+"%"),
		))
	}

	var admins []*Admin
	_, err := builder.LoadContext(ctx, &admins)
	if err != nil {
		b.Log.Panic(err)
	}

	for _, admin := range admins {
		l = append(l, AdminResolver{admin})
	}

	return l
}

// Total admins list total
func (r *AdminListResolver) Total(ctx context.Context) int32 {
	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)
	builder := db.Select("COUNT(*) as count").
		From("admin").
		Where("dtime IS NULL")

	if r.search != "" {
		builder.Where(dbr.Or(
			dbr.Expr("name LIKE ?", "%"+r.search+"%"),
		))
	}

	err := builder.LoadOneContext(ctx, r)
	if err != nil {
		b.Log.Panic(err)
	}

	return r.Count
}

// Admin resolver
func (r *resolver) Admin(
	ctx context.Context,
	args struct{ ID graphql.ID },
) *AdminResolver {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	id := ctx.Value(b.CtxKey("id")).(string)

	if id != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return nil
	}

	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)

	var admin Admin
	err := db.Select("*").
		From("admin").
		Where("dtime IS NULL").
		Where(dbr.Eq("id", args.ID)).
		LoadOneContext(ctx, &admin)
	if err != nil {
		return nil
	}

	return &AdminResolver{&admin}
}

// ID field
func (r *AdminResolver) ID() *graphql.ID {
	return r.admin.ID
}

// Name field
func (r *AdminResolver) Name() *string {
	return r.admin.Name
}

// Ctime field
func (r *AdminResolver) Ctime() *string {
	if r.admin.Ctime.Valid {
		return bt.DateTime(&r.admin.Ctime.Time)
	}
	return nil
}

// Mtime field
func (r *AdminResolver) Mtime() *string {
	if r.admin.Mtime.Valid {
		return bt.DateTime(&r.admin.Mtime.Time)
	}
	return nil
}

// Ltime field
func (r *AdminResolver) Ltime() *string {
	if r.admin.Ltime.Valid {
		return bt.DateTime(&r.admin.Ltime.Time)
	}
	return nil
}

// Btime field
func (r *AdminResolver) Btime() *string {
	if r.admin.Btime.Valid {
		return bt.DateTime(&r.admin.Btime.Time)
	}
	return nil
}

// AdminNameCheck check whether the name is duplicated
func (r *resolver) AdminNameCheck(
	ctx context.Context,
	args struct {
		ID   graphql.ID
		Name string
	},
) bool {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	id := ctx.Value(b.CtxKey("id")).(string)

	if id != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return false
	}

	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)

	var admin Admin
	err := db.Select("id").
		From("admin").
		Where("dtime IS NULL").
		Where(dbr.Neq("id", args.ID)).
		Where(dbr.Eq("name", args.Name)).
		LoadOneContext(ctx, &admin)
	if err != nil {
		return false
	}

	return true
}

// UpsertAdmin to upsert admin
func (r *resolver) UpsertAdmin(ctx context.Context, args struct {
	ID   graphql.ID
	Data AdminInput
}) bool {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	sid := ctx.Value(b.CtxKey("id")).(string)

	if sid != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return false
	}

	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)
	nowTime := time.Now()
	now := bt.DateTime(&nowTime)

	id := string(args.ID)
	if id == "0" {
		if args.Data.Name == nil || args.Data.Passwd == nil {
			b.Log.Panic("name and passwd required")
		}

		var admin Admin
		err := db.Select("*").
			From("admin").
			Where("dtime IS NULL").
			Where(dbr.Eq("name", args.Data.Name)).
			LoadOneContext(ctx, &admin)
		if err == nil {
			b.Log.Panic("name duplicated")
		}

		passwd, err := bcrypt.GenerateFromPassword([]byte(*args.Data.Passwd), 10)
		if err != nil {
			b.Log.Panic(err)
		}

		_, err = db.InsertInto("admin").
			Columns("name", "passwd", "ctime", "mtime").
			Values(args.Data.Name, passwd, now, now).
			ExecContext(ctx)
		if err != nil {
			b.Log.Panic(err)
		}
	} else {
		var admin Admin
		err := db.Select("*").
			From("admin").
			Where("dtime IS NULL").
			Where(dbr.Neq("id", id)).
			Where(dbr.Eq("name", args.Data.Name)).
			LoadOneContext(ctx, &admin)
		if err == nil {
			b.Log.Panic("name duplicated")
		}

		if args.Data.Passwd != nil {
			passwd, err := bcrypt.GenerateFromPassword([]byte(*args.Data.Passwd), 10)
			if err != nil {
				b.Log.Panic(err)
			}

			p := string(passwd)
			args.Data.Passwd = &p
		}

		_, err = db.Update("admin").
			Where(dbr.Eq("id", id)).
			SetMap(structs.Map(args.Data)).
			Set("mtime", now).
			ExecContext(ctx)
		if err != nil {
			b.Log.Panic(err)
		}
	}

	return true
}

// BanAdmins ban admins by a list of id
func (r *resolver) BanAdmins(ctx context.Context, args struct {
	Ids    []graphql.ID
	Status bool
}) bool {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	id := ctx.Value(b.CtxKey("id")).(string)

	if id != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return false
	}

	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)
	nowTime := time.Now()
	now := bt.DateTime(&nowTime)

	var err error
	if args.Status == true {
		_, err = db.Update("admin").
			Where("id IN ?", args.Ids).
			Set("btime", now).
			ExecContext(ctx)
	} else {
		_, err = db.Update("admin").
			Where("id IN ?", args.Ids).
			Set("btime", nil).
			ExecContext(ctx)
	}

	if err != nil {
		return false
	}

	return true
}

// DeleteAdmins delete admins by a list of id
func (r *resolver) DeleteAdmins(ctx context.Context, args struct {
	Ids []graphql.ID
}) bool {
	h := ctx.Value(b.CtxKey("http")).(*b.HTTP)
	id := ctx.Value(b.CtxKey("id")).(string)

	if id != "1" {
		http.Error(h.Response, "Forbidden", http.StatusForbidden)
		return false
	}

	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)
	nowTime := time.Now()
	now := bt.DateTime(&nowTime)

	_, err := db.Update("admin").
		Where("id IN ?", args.Ids).
		Set("dtime", now).
		ExecContext(ctx)
	if err != nil {
		return false
	}

	return true
}

// Profile get yourself's profile
func (r *resolver) Profile(ctx context.Context) *AdminResolver {
	id := ctx.Value(b.CtxKey("id")).(string)
	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)

	var admin Admin
	err := db.Select("*").
		From("admin").
		Where("dtime IS NULL").
		Where(dbr.Eq("id", id)).
		LoadOneContext(ctx, &admin)
	if err != nil {
		return nil
	}

	return &AdminResolver{&admin}
}

// UpdateProfile update yourself's profile
func (r *resolver) UpdateProfile(ctx context.Context, args struct {
	Data *AdminInput
}) bool {
	id := ctx.Value(b.CtxKey("id")).(string)
	db := ctx.Value(b.CtxKey("dbr")).(*dbr.Session)
	nowTime := time.Now()
	now := bt.DateTime(&nowTime)

	var admin Admin
	err := db.Select("*").
		From("admin").
		Where("dtime IS NULL").
		Where(dbr.Neq("id", id)).
		Where(dbr.Eq("name", args.Data.Name)).
		LoadOneContext(ctx, &admin)
	if err == nil {
		b.Log.Panic("name duplicated")
	}

	if args.Data.Passwd != nil {
		passwd, err := bcrypt.GenerateFromPassword([]byte(*args.Data.Passwd), 10)
		if err != nil {
			b.Log.Panic(err)
		}

		p := string(passwd)
		args.Data.Passwd = &p
	}

	_, err = db.Update("admin").
		Where(dbr.Eq("id", id)).
		SetMap(structs.Map(args.Data)).
		Set("mtime", now).
		ExecContext(ctx)
	if err != nil {
		b.Log.Panic(err)
	}

	return true
}
