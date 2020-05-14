# frozen_string_literal: true

class StatusEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :kind, Types::StatusKinds
  attribute? :likes_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :work, WorkEntity
end
