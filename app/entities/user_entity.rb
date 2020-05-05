# frozen_string_literal: true

class UserEntity < ApplicationEntity
  attribute? :username, Types::String
  attribute? :name, Types::String.optional
  attribute? :description, Types::String.optional
  attribute? :avatar_url, Types::String.optional
  attribute? :is_supporter, Types::Bool
end
