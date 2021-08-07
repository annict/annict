# frozen_string_literal: true

# == Schema Information
#
# Table name: flashes
#
#  id          :bigint           not null, primary key
#  client_uuid :string           not null
#  data        :json
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_flashes_on_client_uuid  (client_uuid) UNIQUE
#

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
