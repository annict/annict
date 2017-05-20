# frozen_string_literal: true

class ModelMismatchError < StandardError; end

class ApplicationRecord < ActiveRecord::Base
  self.abstract_class = true

  def self.find_by_graphql_id(graphql_id)
    type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(graphql_id)
    raise ModelMismatchError if Object.const_get(type_name) != self
    find item_id
  end

  def root_resource?
    false
  end
end
