# frozen_string_literal: true

class CastPolicy < ApplicationPolicy
  def create?
    user.committer?
  end

  def update?
    user.committer?
  end

  def hide?
    signed_in? && user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
