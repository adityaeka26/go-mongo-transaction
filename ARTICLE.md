# MongoDB Transaction in Go

MongoDB adalah salah satu database NoSQL paling populer di dunia saat ini. MongoDB merupakan database yang berbentuk dokumen dimana data disimpan dalam belum format yang seperti JSON, yaitu BSON (Binary JSON). Berbeda dengan database SQL, MongoDB menggunakan dokumen yang fleksibel untuk menyimpan datanya sehingga manajemen data lebih dinamis dan scalable karena struktur data dapat ditambah dan dimodifikasi tanpa menentukan skema terlebih dahulu.

Di Telkom Indonesia sendiri, MongoDB sudah sering sekali digunakan di banyak project dan aplikasi karena skalabilitas dan fleksibilitasnya yang sangat baik. Salah satu fitur yang powerful di MongoDB adalah transaction. Di artikel ini, akan dibahas bagaimana cara mengimplementasikan MongoDB transaction di Golang, mengingat saat ini di Telkom Indonesia sudah mulai banyak project yang menggunakan Golang.

## Database Transaction

Database transaction adalah fitur powerful di database yang berfungsi untuk menjaga integritas dan konsistensi data. Transaction memungkinkan kita untuk menyatukan banyak operasi menjadi satu operasi atomik. Dengan fitur transaction ini, semua operasi akan dipastikan berhasil atau tidak sama sekali, sehingga jika ada salah satu operasi yang gagal, maka operasi yang lainnya juga harus gagal.

## MongoDB Transaction

MongoDB memperkenalkan fitur database transaction di versi 4.0 yang dirilis pada tahun 2018 lalu. Database transaction di MongoDB bisa dilakukan untuk semua operasi read dan write (CRUD) mulai dari insert, update, delete, find, aggregate, dan lain-lain.

### Replica Set

Untuk dapat menggunakan fitur transaction di MongoDB, ada syarat yang harus dipenuhi, yaitu **MongoDB harus berjalan di replica set** karena membutuhkan capped collection yang bernama oplog (operations log). Transaction tidak bisa dilakukan di MongoDB standalone karena tidak ada oplog.

Hal ini bukanlah masalah karena di environment production kebanyakan sudah menggunakan replica set MongoDB untuk menjaga high-availability. Untuk development, di artikel ini juga sudah disediakan docker compose untuk menjalankan MongoDB replica set di komputer lokal.

## Implementasi MongoDB Transaction di Go

Package yang digunakan di artikel ini adalah package official MongoDB driver dimana package ini merupakan package rekomendasi dari MongoDB untuk melakukan transaction. API transaction yang digunakan adalah callback API, prosesnya antara lain:
1. Start transaction
2. Execute operations (tidak harus operasi MongoDB, bisa operaasi lainnya)
3. Jika seluruh operasi berhasil, commit transaction. Namun jika ada salah satu operasi yang gagal, abort transaction.

Berikut ini contoh kode transaction MongoDB di Go
```go
app.Get("/transaction/success", func(c *fiber.Ctx) error {
	ctx := c.UserContext()

	session, err := mongoClient.StartSession()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"success": false,
			"data":    err.Error(),
		})
	}
	defer session.EndSession(ctx)

	callback := func(sessionContext mongo.SessionContext) (any, error) {
		ctx := mongo.NewSessionContext(ctx, session)

		password, err := bcrypt.GenerateFromPassword([]byte("goat"), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		_, err = mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
			"name":        "Leo Messi",
			"age":         "34",
			"nationality": "Argentina",
			"password":    string(password),
		})
		if err != nil {
			return nil, err
		}

		password, err = bcrypt.GenerateFromPassword([]byte("brazil"), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		_, err = mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
			"name":        "Neymar",
			"age":         "31",
			"nationality": "Brazil",
			"password":    string(password),
		})
		if err != nil {
			return nil, err
		}

		return "transaction success", nil
	}

	result, err := session.WithTransaction(
		ctx,
		callback,
	)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"success": false,
			"data":    err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"success": true,
		"data":    result,
	})
})
```
Keterangan kode:
- Kode di atas merupakan contoh implementasi MongoDB transaction di REST API Go Fiber.
- `ctx := c.UserContext()` adalah function bawaan dari Go Fiber untuk mendapatkan context dari request user. Context ini diperlukan untuk membuat session transaction.
- Ada 4 operasi yang dilakukan di kode di atas, yaitu 2 kali generate password dengan bcrypt dan 2 kali insert data ke MongoDB.
- Seluruh operasi dilakukan di callback, jika ada operasi yang gagal dan return error di callback, maka seluruh operasi MongoDB akan dibatalkan.

Untuk mensimulasikan error di transaction, kita bisa membuat salah satu operasi menjadi error. Contoh:
```go
password, err = bcrypt.GenerateFromPassword([]byte("brazil"), bcrypt.DefaultCost)
err = errors.New("error simulation")
if err != nil {
	return nil, err
}
```
Sebelum kode tersebut, ada 2 operasi yang dilakukan, yaitu generate password dan insert data pertama. Jika tanpa menggunakan transaction maka operasi insert data Leo Messi tetap berhasil dilakukan. Karena kita menggunakan transaction, maka insert data Leo Messi akan dibatalkan.

Kode lengkap bisa dilihat di repository GitHub berikut: https://github.com/adityaeka26/go-mongo-transaction.

Dec 1, 2023<br>
Aditya Eka Bagaskara