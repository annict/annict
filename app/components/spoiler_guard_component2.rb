# frozen_string_literal: true

class SpoilerGuardComponent2 < ApplicationComponent
  def initialize(record:, current_user: nil)
    @record = record
    @current_user = current_user
  end

  def init_is_spoiler
    return false unless @current_user

    @current_user.hide_record_body? && @record.is_spoiler
  end
end
