# Service de gestion de tâches avec authentification basique

Le but est de réaliser un service web pour gérer des tâches. Avec la gestion de l'autentification d'utilisateurs.

Ecrire votre nom et email en commentaire en entête de tout les fichiers .go

## Architecture

Le service est composé de routes d'authentification :

- POST /user/signup
- POST /user/login
- GET /user/logout

Les routes ci dessous devrons être authentifié via un middleware:

- GET /user
- POST /todos
- GET /todos  // la list des todos du user
- GET /todos/:id
- UPDATE /todos/:id
- DELETE /todos/:id

Les données des todos et users seront stoquées dans une base de donnée `PostgreSQL`, les tokens de session dans une base de donnée `Redis`.

Des interfaces `*Store` sont déjà défini ainsi qu'une struct `Service` qui accueillera les implémentations à ces interfaces.
Vos `gin.Handler` devrons être des méthodes de `Service`, pour pouvoir avoir accès au différent `Store`.

Les tokens de session seront stoqués dans un cookie. Un tokens est une string random de 20 caractères à générer au `signup` ou `login`.

## Package à utiliser

Ils sont déjà présent dans le go.mod fourni

- github.com/joho/godotenv
- github.com/gin-gonic/gin
- github.com/gofrs/uuid
- github.com/go-redis/redis/v8
- github.com/jackc/pgx/v4/pgxpool

## Utilisation de docker

Un fichier `docker-compose.yml` est fourni vous pouvez le modifier si besoin.

Pour lancer PostgreSQL et Redis

```sh
docker-compose up
```

Pour supprimer toutes les données et repartir à zero :

```sh
docker-compose rm -fvs
```

Pour récupérer les paramètres de connexion les changer dans `.env` si besoin.

## Données

Les id sont de type `uuid.UUID` en générer des `V4`.

### User

```json
{
    "id": "1179459b-8a7c-4c38-b0ab-433d7ec3b958",
    "name": "myname"
}
```

### Todo

```json
{
    "id": "ffd430e8-a23b-4a02-80d1-73b07b265712",
    "text": "Faire le tp de go",
    "done": false,
    "user_id": "1179459b-8a7c-4c38-b0ab-433d7ec3b958"
}
```

### Session 

Un userid en clé et un token en valeur.
Utiliser les commandes redis `Set`, `Get` et `Del`.

## Routes

Les routes http sont à ajouter à la méthode `SetupRoute` de `Service`

Toute les routes doivent répondre avec un json `{ "error": "message" } ` en cas d'erreur `{ "data" : ... }` en cas de succes avec un status http adéquat.

Si vous n'arrivez pas implementer des méthodes store faite la retourner une erreur ou une valeur par défaut, pour que vos struct store satisface leurs interfaces et ne pas être bloqué.

```go
return nil, fmt.Errorf("unimplemented")
```

## Middleware d'authentification

Créer un middleware `gin` qui va vérifier la présence d'un cookie et vérifier ca validité via redis.
S'il n'est pas valide retourner un status `401` et un message d'erreur. Affecter le userid obtenue dans le context de la connection gin.

Les handler pouront alors récupérer via le context gin le userid du compte connecté.
