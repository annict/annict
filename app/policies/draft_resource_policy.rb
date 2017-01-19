class DraftResourcePolicy < ApplicationPolicy
  def create?
    banned_ids = ENV.fetch("ANNICT_BANNED_WRITING_USER_IDS").
      split(",").
      map(&:strip).
      map(&:to_i)
    user.present? && !user.id.in?(banned_ids)
  end

  def update?
    user.present? &&
      user == record.edit_request.user &&
      !record.edit_request.published? &&
      !record.edit_request.closed?
  end
end
