# typed: false
# frozen_string_literal: true

class SeriesWorkPolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def update?
    user.present? && user.committer?
  end

  def destroy?
    user.present? && user.admin?
  end

  def unpublish?
    user.present? && user.committer?
  end
end
