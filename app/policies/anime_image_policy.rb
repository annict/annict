# frozen_string_literal: true

class AnimeImagePolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def update?
    user.present? && user.committer?
  end
end
