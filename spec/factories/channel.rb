FactoryGirl.define do
  factory :channel do
    association :channel_group
    sequence(:sc_chid)
    name 'テレビ夕日'
  end
end