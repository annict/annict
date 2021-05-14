# frozen_string_literal: true

module Deprecated
  class OrganizationEntity < Deprecated::ApplicationEntity
    local_attributes :name

    attribute? :database_id, Types::Integer
    attribute? :name, Types::String
    attribute? :name_en, Types::String.optional

    def self.from_node(organization_node)
      attrs = {}

      if database_id = organization_node["databaseId"]
        attrs[:database_id] = database_id
      end

      if name = organization_node["name"]
        attrs[:name] = name
      end

      if name_en = organization_node["nameEn"]
        attrs[:name_en] = name_en
      end

      new attrs
    end
  end
end
