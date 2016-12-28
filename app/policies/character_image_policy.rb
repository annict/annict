# frozen_string_literal: true

class CharacterImagePolicy < ApplicationPolicy
  def destroy?
    user.present? && user.id == record.user.id
  end
end
