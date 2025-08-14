# Example usage of tama_thought_initializer resource

resource "tama_space" "example" {
  name = "example-space"
  type = "root"
}

resource "tama_chain" "example" {
  space_id = tama_space.example.id
  name     = "Example Processing Chain"
}

resource "tama_class" "movie_details" {
  space_id = tama_space.example.id
  schema_json = jsonencode({
    title       = "Movie Details Schema"
    description = "Schema for movie details"
    type        = "object"
    properties = {
      title = {
        type        = "string"
        description = "Movie title"
      }
      description = {
        type        = "string"
        description = "Movie description"
      }
      overview = {
        type        = "string"
        description = "Movie overview"
      }
      setting = {
        type        = "string"
        description = "Movie setting"
      }
    }
    required = ["title"]
  })
}

resource "tama_modular_thought" "some_thought" {
  chain_id        = tama_chain.example.id
  output_class_id = tama_class.movie_details.id
  relation        = "preload"

  module {
    reference = "tama/agentic/generate"
    parameters = jsonencode({
      relation = "preload"
    })
  }
}

resource "tama_thought_initializer" "preload_data" {
  thought_id = tama_modular_thought.some_thought.id

  reference = "tama/initializers/preload"
  index     = 0
  class_id  = tama_class.movie_details.id
  parameters = jsonencode({
    concept = {
      relations  = ["description", "overview", "setting"]
      embeddings = "include"
      content = {
        action = "merge"
        merge = {
          location = "root"
        }
      }
    }
    children = [
      {
        class = "movie-credits"
        as    = "object"
      }
    ]
  })
}
