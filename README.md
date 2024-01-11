# Esqlo

Build data-focused UIs fast with just SQL and HTML.

```html
<sql src="postgres://localhost:12345" id="users">
  SELECT id, name, account_age FROM users LIMIT 10
</sql>

{{#each users}}
  <div>
    <h1>{{name}}</h1>
    <p>Account age: {{account_age}}</p>
  </div>
{{/each}}
```

# Installation
### Build from source

Requirements:
- Go 1.20 or higher
- CGO enabled

```shell
go build -o bin/esqlo cmd/esqlo/main.go

./bin/esqlo -listen 127.0.0.1:8080 -serve static/
```

TODO: Add prebuilt binaries download link

# Usage
Create or use an existing database file. A database can be anything from a JSON file to a MySQL database. Anything that
cann accept SQL queries and returns JSON is usable!

Let's start by creating a file called `static/reviews.csv` with some data in it:
```csv
# static/reviews.csv
id,restauarant,reviewer,stars,review
1,McDonalds,"John Doe",5,"I love McDonalds!"
2,McDonalds,"Jane Doe",1,"I hate McDonalds!"
3,McDonalds,Blake,3,"I just wanted Burger King."
```

Now that we have data, we want to make a UI to display it. We can just write normal HTML.

```html
<!-- static/index.html -->
<h1>Reviews</h1>
<table>
  <tr>
    <th>Restaurant</th>
    <th>Reviewer</th>
    <th>Stars</th>
    <th>Review</th>
  </tr>
  <tr>
    <td>McDonalds</td>
    <td>John Doe</td>
    <td>5</td>
    <td>I love McDonalds!</td>
  </tr> 
</table>
```

that looks okay, but we'd like it to automatically use our data. We can do that by adding a `<sql>` tag to our HTML using the DuckDB database to load everything for us automatically.

```html
<!-- static/index.html -->
<h1>Reviews</h1>
<table>
  <tr>
    <th>Restaurant</th>
    <th>Reviewer</th>
    <th>Stars</th>
    <th>Review</th>
  </tr>
  <!-- Run an SQL query against our local file using duckdb, and save the result in a table accessible in this HTML file -->
  <sql src="duckdb" id="reviews">
    SELECT restaurant, reviewer, stars, review FROM "static/reviews.csv"
  </sql>

  <!-- This is Mustache syntax. Here, we are saying "go over every review, and replace {{xxx}} with the result from our SQL query -->
  {{#reviews}}
    <tr>
      <td>{{restaurant}}</td>
      <td>{{reviewer}}</td>
      <td>{{stars}}</td>
      <td>{{review}}</td>
    </tr>
  {{/reviews}}
</table>
```

Now, we can run esqlo and see the result:

```shell
ls static/
# index.html style.css index.js ...

# Run your web server on port 8080 serving the local static/ directory
esqlo -listen 127.0.0.1:8080 -serve static/
```

Click on http://localhost:8080 and you should see the data from CSV file. If you modify the HTML file or change the database, the data updates on reload.

### Databases supported
- [x] DuckDB (local CSV, JSON, Parquet as well)
- [ ] MySQL
- [ ] PostgreSQL
- [ ] SQLite
- [ ] MongoDB
- [ ] Redis
- [ ] Athena

### Features
- [x] Run SQL queries against any database and use the results in your HTML
- [ ] Parameterized pages and queries using query parameters (e.g. `?id=123` can be referenced as `{{query.id}}`)
- [ ] Caching mechanism for SQL queries
- [ ] Websocket support for realtime updates (<100 ms per update)
- [ ] JSON rendering of `<sql>` tags for easier use in Javascript


