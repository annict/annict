# frozen_string_literal: true

class UserEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :username, Types::String
  attribute? :name, Types::String.optional
  attribute? :description, Types::String.optional
  attribute? :avatar_url, Types::String.optional
  attribute? :background_image_url, Types::String.optional
  attribute? :display_supporter_badge, Types::Bool
end
