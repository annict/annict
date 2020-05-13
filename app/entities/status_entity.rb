# frozen_string_literal: true

class StatusEntity < ApplicationEntity
  attribute :type, Types::Value("status")

  attribute? :id, Types::Integer
  attribute? :kind, Types::StatusKinds
  attribute? :likes_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :work, WorkEntity
end
