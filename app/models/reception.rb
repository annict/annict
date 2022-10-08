# == Schema Information
#
# Table name: receptions
#
#  id         :bigint           not null, primary key
#  created_at :timestamptz
#  updated_at :timestamptz
#  channel_id :bigint           not null
#  user_id    :bigint           not null
#
# Indexes
#
#  receptions_channel_id_idx          (channel_id)
#  receptions_user_id_channel_id_key  (user_id,channel_id) UNIQUE
#  receptions_user_id_idx             (user_id)
#
# Foreign Keys
#
#  receptions_channel_id_fk  (channel_id => channels.id) ON DELETE => cascade
#  receptions_user_id_fk     (user_id => users.id) ON DELETE => cascade
#

class Reception < ApplicationRecord
  belongs_to :channel
  belongs_to :user

  def self.initial?(reception)
    count == 1 && first.id == reception.id
  end
end
