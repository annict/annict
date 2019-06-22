# typed: false
# frozen_string_literal: true

class ChangeSchemaForV3 < ActiveRecord::Migration[6.1]
  def change
    enable_extension(:pgcrypto) unless extension_enabled?("pgcrypto")
    enable_extension(:citext) unless extension_enabled?("citext")
  end
end
