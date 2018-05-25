# frozen_string_literal: true

class RecordPolicy < ApplicationPolicy
  def destroy?
    user == record.user
  end
end
