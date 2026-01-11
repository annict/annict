# typed: false
# frozen_string_literal: true

describe User, type: :model do
  describe "#supporter?" do
    context "Stripeサブスクリプションがアクティブな場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :active) }
      let(:user) { create(:user, stripe_subscriber: stripe_subscriber) }

      it "trueを返す" do
        expect(user.supporter?).to eq true
      end
    end

    context "StripeサブスクリプションがPast Dueの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :past_due) }
      let(:user) { create(:user, stripe_subscriber: stripe_subscriber) }

      it "trueを返す（支払い遅延中も猶予期間として利用可能）" do
        expect(user.supporter?).to eq true
      end
    end

    context "Stripeサブスクリプションがキャンセル済みの場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :canceled) }
      let(:user) { create(:user, stripe_subscriber: stripe_subscriber) }

      it "falseを返す" do
        expect(user.supporter?).to eq false
      end
    end

    context "Gumroadサブスクリプションがアクティブな場合" do
      let(:gumroad_subscriber) { create(:gumroad_subscriber) }
      let(:user) { create(:user, gumroad_subscriber: gumroad_subscriber) }

      it "trueを返す" do
        expect(user.supporter?).to eq true
      end
    end

    context "Gumroadサブスクリプションがキャンセル済みの場合" do
      let(:gumroad_subscriber) { create(:gumroad_subscriber, gumroad_cancelled_at: 1.day.ago) }
      let(:user) { create(:user, gumroad_subscriber: gumroad_subscriber) }

      it "falseを返す" do
        expect(user.supporter?).to eq false
      end
    end

    context "StripeとGumroad両方がアクティブな場合" do
      let(:stripe_subscriber) { create(:stripe_subscriber, :active) }
      let(:gumroad_subscriber) { create(:gumroad_subscriber) }
      let(:user) { create(:user, stripe_subscriber: stripe_subscriber, gumroad_subscriber: gumroad_subscriber) }

      it "trueを返す" do
        expect(user.supporter?).to eq true
      end
    end

    context "どちらのサブスクリプションもない場合" do
      let(:user) { create(:user) }

      it "falseを返す" do
        expect(user.supporter?).to eq false
      end
    end
  end

  describe "#create_or_last_activity_group!" do
    context "when itemable is Status object" do
      let(:user) { create :user }
      let(:status) { create(:status, user: user) }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 0

        user.create_or_last_activity_group!(status)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq "Status"
        expect(activity_group.single).to eq false
      end
    end

    context "when itemable is EpisodeRecord object and it has body" do
      let(:user) { create :user }
      let(:episode_record) { create(:episode_record, user: user, body: "良かった") }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 0

        user.create_or_last_activity_group!(episode_record)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq "EpisodeRecord"
        expect(activity_group.single).to eq true
      end
    end

    context "when itemable is EpisodeRecord object and it does not have body" do
      let(:user) { create :user }
      let(:episode_record) { create(:episode_record, user: user, body: "") }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 0

        user.create_or_last_activity_group!(episode_record)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq "EpisodeRecord"
        expect(activity_group.single).to eq false
      end
    end

    context "when itemable is WorkRecord object and it has body" do
      let(:user) { create :user }
      let(:work_record) { create(:work_record, user: user, body: "良かった") }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 0

        user.create_or_last_activity_group!(work_record)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq "WorkRecord"
        expect(activity_group.single).to eq true
      end
    end

    context "when itemable is WorkRecord object and it does not have body" do
      let(:user) { create :user }
      let(:work_record) { create(:work_record, user: user, body: "") }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 0

        user.create_or_last_activity_group!(work_record)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq "WorkRecord"
        expect(activity_group.single).to eq false
      end
    end

    context "when activity group which itemable_type is same and not single is created" do
      let(:user) { create :user }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Status", single: false) }
      let(:status) { create(:status, user: user) }

      it "does not create activity group" do
        expect(ActivityGroup.count).to eq 1

        user.create_or_last_activity_group!(status)

        expect(ActivityGroup.count).to eq 1

        activity_group = user.activity_groups.first

        expect(activity_group.itemable_type).to eq activity_group.itemable_type
        expect(activity_group.single).to eq activity_group.single
      end
    end

    context "when activity group which itemable_type is same and not single is created but it created more than 12 hours ago" do
      let(:user) { create :user }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Status", single: false, created_at: 13.hour.ago) }
      let(:status) { create(:status, user: user) }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 1

        user.create_or_last_activity_group!(status)

        expect(ActivityGroup.count).to eq 2

        activity_group = user.activity_groups.last

        expect(activity_group.itemable_type).to eq "Status"
        expect(activity_group.single).to eq false
      end
    end

    context "when activity group which itemable_type is not same is created" do
      let(:user) { create :user }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: false) }
      let(:status) { create(:status, user: user) }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 1

        user.create_or_last_activity_group!(status)

        expect(ActivityGroup.count).to eq 2

        activity_group = user.activity_groups.last

        expect(activity_group.itemable_type).to eq "Status"
        expect(activity_group.single).to eq false
      end
    end

    context "when activity group which itemable_type is same and single is created" do
      let(:user) { create :user }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
      let(:episode_record) { create(:episode_record, user: user, body: "良かった") }

      it "creates activity group" do
        expect(ActivityGroup.count).to eq 1

        user.create_or_last_activity_group!(episode_record)

        expect(ActivityGroup.count).to eq 2

        activity_group = user.activity_groups.last

        expect(activity_group.itemable_type).to eq "EpisodeRecord"
        expect(activity_group.single).to eq true
      end
    end
  end
end
