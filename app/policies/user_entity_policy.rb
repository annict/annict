# frozen_string_literal: true

class UserEntityPolicy < ApplicationPolicy
  def mute?
    user&.id != record.database_id
  end
end
