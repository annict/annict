# typed: false
# frozen_string_literal: true

class Flash < ApplicationRecord
  def self.store_data(client_uuid, hash)
    return if client_uuid.blank?
    flash = Flash.where(client_uuid: client_uuid).first_or_create
    flash.update_column(:data, {type: hash.keys.first, message: hash.values.first})
  end

  def self.reset_data(client_uuid)
    flash = Flash.find_by(client_uuid: client_uuid)
    return if flash.nil?
    flash.update_column(:data, {})
  end
end
