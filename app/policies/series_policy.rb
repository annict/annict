# frozen_string_literal: true

class SeriesPolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def update?
    user.present? && user.committer?
  end

  def publish?
    user.present? && user.committer?
  end

  def destroy?
    user.present? && user.role.admin?
  end
end
