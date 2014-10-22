class Tip < ActiveRecord::Base
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
end
