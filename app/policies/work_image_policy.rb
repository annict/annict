# typed: false
# frozen_string_literal: true

class WorkImagePolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def update?
    user.present? && user.committer?
  end
end
