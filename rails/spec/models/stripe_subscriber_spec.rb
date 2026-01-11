# typed: false
# frozen_string_literal: true

describe StripeSubscriber, type: :model do
  describe "#active?" do
    context "stripe_statusがactiveの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :active) }

      it "trueを返す" do
        expect(stripe_subscriber.active?).to eq true
      end
    end

    context "stripe_statusがpast_dueの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :past_due) }

      it "trueを返す（支払い遅延中も猶予期間として利用可能）" do
        expect(stripe_subscriber.active?).to eq true
      end
    end

    context "stripe_statusがcanceledの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :canceled) }

      it "falseを返す" do
        expect(stripe_subscriber.active?).to eq false
      end
    end

    context "stripe_statusがunpaidの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :unpaid) }

      it "falseを返す" do
        expect(stripe_subscriber.active?).to eq false
      end
    end
  end
end
