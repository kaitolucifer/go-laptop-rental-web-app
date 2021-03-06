# Go Laptop Rental Service

<pre>
            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
                    Version 2, December 2004

 Copyright (C) 2004 Sam Hocevar <sam@hocevar.net>

 Everyone is permitted to copy and distribute verbatim or modified
 copies of this license document, and changing it is allowed as long
 as the name is changed.

            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
   TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION

  0. You just DO WHAT THE FUCK YOU WANT TO.
</pre>

A web application of laptop rental service using Go.

- Go
  - Built in Go version 1.16
  - Uses the [chi router](https://github.com/go-chi/chi)
  - Uses alex edwards [SCS session management](https://github.com/alexedwards/scs)
  - Uses [nosurf](https://github.com/justinas/nosurf)
  - Uses [govalidator](https://github.com/asaskevich/govalidator)
  - Uses [pop database toolkit](https://github.com/gobuffalo/pop)
  - Uses [go-simple-mail](https://github.com/xhit/go-simple-mail) for sending Email
  - Uses [GoDotEnv](https://github.com/joho/godotenv)
- HTML / CSS / JavaScript
  - Uses [Bootstrap 5](https://getbootstrap.jp/)
  - Uses [RoyalUI-Free-Bootstrap-Admin-Template](https://github.com/BootstrapDash/RoyalUI-Free-Bootstrap-Admin-Template)
  - Uses [Foundation for Emails 2](https://get.foundation/emails.html)
  - Uses [vanillajs-datepicker](https://github.com/mymth/vanillajs-datepicker)
  - Uses the [notie notification suite](https://github.com/jaredreich/notie)
  - Uses the [sweetalert2 modal](https://sweetalert2.github.io/)
  - Uses [Simple-DataTables](https://github.com/fiduswriter/Simple-DataTables)

### Screeshots

- Home Page
  ![home](https://github.com/kaitolucifer/go-laptop-rental-web-app/blob/main/demo/home.png)

- Reservation
  ![reservation](https://github.com/kaitolucifer/go-laptop-rental-web-app/blob/main/demo/reservation.png)

- Admin Page
  ![admin](https://github.com/kaitolucifer/go-laptop-rental-web-app/blob/main/demo/admin.png)

### Installation

- `go get` to download all modules
- `cd dockerfile && docker-compose up -d` to start postgresql and mailhog service
- fill `database.yml` and `.env` with information in `dockerfile/docker-compose.yml`
- install [pop database toolkit](https://github.com/gobuffalo/pop)
- `soda migrate` to migrate database
- `chmod +x run.sh && ./run.sh` to run server
- default admin email and password
  - email: `admin@admin.com`
  - password: `password`
