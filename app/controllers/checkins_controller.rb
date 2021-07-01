# frozen_string_literal: true

class CheckinsController < ApplicationV6Controller
  # Old record page
  def show
    record = Record.only_kept.find(params[:id])
    redirect_to record_path(record.user.username, record), status: 301
  end
end
