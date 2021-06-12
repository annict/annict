# frozen_string_literal: true

class PersonEntity < ApplicationEntity
  local_attributes :name

  attribute? :database_id, Types::Integer
  attribute? :name, Types::String
  attribute? :name_en, Types::String.optional

  def self.from_node(person_node)
    attrs = {}

    if (database_id = person_node["databaseId"])
      attrs[:database_id] = database_id
    end

    if (name = person_node["name"])
      attrs[:name] = name
    end

    if (name_en = person_node["nameEn"])
      attrs[:name_en] = name_en
    end

    new attrs
  end
end
