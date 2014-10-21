# == Schema Information
#
# Table name: tips
#
#  id           :integer          not null, primary key
#  title        :string(255)      not null
#  partial_name :string(255)      not null
#  target       :integer          not null
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#

class Tip < ActiveRecord::Base
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
end
