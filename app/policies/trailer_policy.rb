# frozen_string_literal: true

class TrailerPolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def update?
    user.present? && user.committer?
  end

  def hide?
    user.present? && user.role.admin?
  end

  def destroy?
    user.present? && user.role.admin?
  end
end
