# frozen_string_literal: true

class WorkImagePolicy < ApplicationPolicy
  def create?
    return false
    user.committer?
  end

  def update?
    return false
    user.committer?
  end
end
