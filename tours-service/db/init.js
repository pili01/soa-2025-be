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


console.log(`Checking for 'keypoints' collection in database: ${process.env.MONGO_INITDB_DATABASE}`);

if (collectionNames.includes('keypoints')) {
    console.log("'keypoints' collection already exists. Skipping creation.");
} else {
    console.log("'keypoints' collection does not exist. Creating now...");
    db.createCollection('keypoints', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "tourId", "name", "latitude", "longitude", "ordinal"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            tourId: {
              bsonType: "int",
              description: "must be an integer and is required. References the tour this keypoint belongs to."
            },
            name: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            description: {
              bsonType: "string",
              description: "must be a string if it exists."
            },
            imageUrl: {
              bsonType: "string",
              description: "must be a string if it exists."
            },
            latitude: {
              bsonType: "double",
              description: "must be a double and is required. Geographic latitude coordinate."
            },
            longitude: {
              bsonType: "double",
              description: "must be a double and is required. Geographic longitude coordinate."
            },
            ordinal: {
              bsonType: "int",
              description: "must be an integer and is required. Order of the keypoint in the tour."
            }
          }
        }
      }
    });

    console.log("Collection 'keypoints' created with validation rules.");

    
    db.keypoints.createIndex({ "tourId": 1 });
    db.keypoints.createIndex({ "tourId": 1, "ordinal": 1 });
    db.keypoints.createIndex({ "latitude": 1, "longitude": 1 });

    console.log("Indexes for 'keypoints' collection created/ensured.");
}

const currentCollections = db.getCollectionNames();

if (currentCollections.includes('tour_reviews')) {
    console.log("'tour_reviews' collection already exists. Skipping creation.");
} else {
    console.log("'tour_reviews' collection does not exist. Creating now...");
    db.createCollection('tour_reviews', {
      validator: {
        $jsonSchema: {
          bsonType: "object",
          required: ["_id", "tourId", "touristId", "rating", "comment", "visitDate", "commentDate"],
          properties: {
            _id: {
              bsonType: "int",
              description: "must be an integer and is required. Managed by the application."
            },
            tourId: {
              bsonType: "int",
              description: "must be an integer and is required. References the tour."
            },
            touristId: {
              bsonType: "int",
              description: "must be an integer and is required. References the tourist who left the review."
            },
            rating: {
              bsonType: "int",
              minimum: 1,
              maximum: 5,
              description: "must be an integer between 1 and 5 and is required."
            },
            comment: {
              bsonType: "string",
              description: "must be a string and is required."
            },
            visitDate: {
              bsonType: "date",
              description: "must be a date and is required. The date the tourist visited the tour."
            },
            commentDate: {
              bsonType: "date",
              description: "must be a date and is required. The date the review was created."
            },
            imageUrls: {
              bsonType: "array",
              description: "must be an array of strings if it exists.",
              items: {
                bsonType: "string"
              }
            }
          }
        }
      }
    });

    console.log("Collection 'tour_reviews' created with validation rules.");

    db.tour_reviews.createIndex({ "tourId": 1 });
    
    db.tour_reviews.createIndex({ "touristId": 1 });

    db.tour_reviews.createIndex({ "tourId": 1, "touristId": 1 }, { unique: true });


    console.log("Indexes for 'tour_reviews' collection created/ensured.");
}

console.log("Database initialization script finished.");