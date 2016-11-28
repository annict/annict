# frozen_string_literal: true
# == Schema Information
#
# Table name: dislikes
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_type :string           not null
#  recipient_id   :integer          not null
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#
# Indexes
#
#  index_dislikes_on_recipient_id_and_recipient_type  (recipient_id,recipient_type)
#  index_dislikes_on_recipient_type_and_recipient_id  (recipient_type,recipient_id)
#  index_dislikes_on_user_id                          (user_id)
#

class Dislike < ApplicationRecord
  belongs_to :recipient, polymorphic: true, counter_cache: true
  belongs_to :user
end
