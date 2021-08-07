# frozen_string_literal: true

class RecordEntityPolicy < ApplicationPolicy
  def update?
    user&.id == record.user.database_id
  end

  def destroy?
    update?
  end
end
