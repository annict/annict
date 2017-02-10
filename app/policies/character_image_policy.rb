# frozen_string_literal: true

class CharacterImagePolicy < ApplicationPolicy
  def create?
    user.committer?
  end

  def update?
    user.committer?
  end
end
