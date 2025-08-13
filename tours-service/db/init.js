db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE);

console.log(`Checking for 'tours' collection in database: ${process.env.MONGO_INITDB_DATABASE}`);

const collectionNames = db.getCollectionNames();

if (collectionNames.includes('tours')) {
    console.log("'tours' collection already exists. Skipping creation.");
} else {
    console.log("'tours' collection does not exist. Creating now...");
    db.createCollection('tours', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "authorId", "name", "description", "difficulty", "status", "price"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            authorId: {
              bsonType: "int",
              description: "must be an integer and is required."
            },
            name: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            description: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            difficulty: {
              enum: ["Easy", "Medium", "Hard"],
              description: "can only be one of the enum values and is required."
            },
            tags: {
              bsonType: "array",
              description: "must be an array of strings if it exists.",
              items: {
                bsonType: "string"
              }
            },
            status: {
              enum: ["Draft", "Published", "Archived"],
              description: "can only be one of the enum values and is required."
            },
            price: {
              bsonType: "double",
              description: "must be a double and is required."
            }
          }
        }
      }
    });

    console.log("Collection 'tours' created with validation rules.");

    db.tours.createIndex({ "authorId": 1 });
    db.tours.createIndex({ "authorId": 1, "name": 1 }, { unique: true });

    console.log("Indexes for 'tours' collection created/ensured.");
}

console.log("Database initialization script finished.");