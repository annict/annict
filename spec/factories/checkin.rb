FactoryGirl.define do
  factory :checkin do
    association :user
    comment "おもしろかった"
    twitter_url_hash "xxxxx"
    episode
    rating 3.0

    before(:create) { Tip.create_with(attributes_for(:checkin_tip)).find_or_create_by(partial_name: 'checkin') }
    before(:create) { |c| c.work = c.episode.work }
  end
end
