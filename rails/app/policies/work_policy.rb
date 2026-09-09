# typed: false
# frozen_string_literal: true

class WorkPolicy < ApplicationPolicy
  def update?
    user.present? && user.committer?
  end

  def destroy?
    user.present? && user.admin? && record.not_deleted?
  end

  def unpublish?
    user.present? && user.committer? && record.not_deleted?
  end
end
