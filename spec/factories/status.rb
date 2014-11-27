FactoryGirl.define do
  factory :status do
    association :user
    association :work
    kind :watching

    before(:create) { Tip.create_with(attributes_for(:status_tip)).find_or_create_by(partial_name: 'status') }
  end
end
