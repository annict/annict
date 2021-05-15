# frozen_string_literal: true

module V4
  class CastEntity < V4::ApplicationEntity
    local_attributes :accurate_name

    attribute? :accurate_name, Types::String
    attribute? :accurate_name_en, Types::String.optional
    attribute? :character, V4::CharacterEntity
    attribute? :person, V4::PersonEntity

    def self.from_nodes(cast_nodes)
      cast_nodes.map do |cast_node|
        from_node(cast_node)
      end
    end

    def self.from_node(cast_node)
      attrs = {}

      if (accurate_name = cast_node["accurateName"])
        attrs[:accurate_name] = accurate_name
      end

      if (accurate_name_en = cast_node["accurateNameEn"])
        attrs[:accurate_name_en] = accurate_name_en
      end

      if (character_node = cast_node["character"])
        attrs[:character] = V4::CharacterEntity.from_node(character_node)
      end

      if (person_node = cast_node["person"])
        attrs[:person] = V4::PersonEntity.from_node(person_node)
      end

      new attrs
    end
  end
end
