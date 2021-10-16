# frozen_string_literal: true

namespace :supporters do
  task sync_with_gumroad: :environment do
    User.only_kept.where.not(gumroad_subscriber_id: nil).preload(:gumroad_subscriber).find_each do |user|
      form = Forms::SupporterForm.new(gumroad_subscriber_id: user.gumroad_subscriber.gumroad_id)

      Updaters::SupporterUpdater.new(user: user, form: form).call
    end
  end
end
