# frozen_string_literal: true

class WorkImagePolicy < ApplicationPolicy
  def create?
    user.committer?
  end

  def update?
    user.committer?
  end
end
