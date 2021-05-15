# frozen_string_literal: true

module V4
  class UserEntityPolicy < ApplicationPolicy
    def mute?
      user&.id != record.database_id
    end
  end
end
