# frozen_string_literal: true

module Deprecated
  class StaffEntity < Deprecated::ApplicationEntity
    local_attributes :accurate_name, :role

    attribute? :accurate_name, Types::String
    attribute? :accurate_name_en, Types::String.optional
    attribute? :role, Types::String
    attribute? :role_en, Types::String.optional
    attribute? :resource do
      attribute :typename, Types::String
      attribute :database_id, Types::Integer
    end

    def self.from_nodes(staff_nodes)
      staff_nodes.map do |staff_node|
        from_node(staff_node)
      end
    end

    def self.from_node(staff_node)
      attrs = {}

      if accurate_name = staff_node["accurateName"]
        attrs[:accurate_name] = accurate_name
      end

      if accurate_name_en = staff_node["accurateNameEn"]
        attrs[:accurate_name_en] = accurate_name_en
      end

      if role = staff_node["role"]
        attrs[:role] = role
      end

      if role_en = staff_node["roleEn"]
        attrs[:role_en] = role_en
      end

      if resource_node = staff_node["resource"]
        attrs[:resource] = {
          typename: resource_node["__typename"],
          database_id: resource_node["databaseId"]
        }
      end

      new attrs
    end

    def person?
      resource.typename == "Person"
    end
  end
end
