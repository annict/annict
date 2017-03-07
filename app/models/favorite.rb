# frozen_string_literal: true
# == Schema Information
#
# Table name: favorites
#
#  id            :integer          not null, primary key
#  user_id       :integer          not null
#  resource_type :string           not null
#  resource_id   :integer          not null
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_favorites_on_user_id                                    (user_id)
#  index_favorites_on_user_id_and_resource_type_and_resource_id  (user_id,resource_type,resource_id) UNIQUE
#

class Favorite < ApplicationRecord
end
