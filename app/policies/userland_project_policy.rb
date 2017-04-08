# frozen_string_literal: true

class UserlandProjectPolicy < ApplicationPolicy
  def update?
     record.users.exists?(user)
  end

  def destroy?
    record.users.exists?(user)
  end
end
