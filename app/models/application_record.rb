# frozen_string_literal: true

class ApplicationRecord < ActiveRecord::Base
  self.abstract_class = true

  LOCALES = %i(en ja other).freeze

  def self.find_by_graphql_id(graphql_id)
    type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(graphql_id)
    raise Annict::Errors::ModelMismatchError if Object.const_get(type_name) != self
    find item_id
  end

  def root_resource?
    false
  end
end
