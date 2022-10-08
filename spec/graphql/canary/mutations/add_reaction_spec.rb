# frozen_string_literal: true

xdescribe Canary::Mutations::AddReaction do
  let(:user_1) { create :registered_user }
  let(:user_2) { create :registered_user }
  let(:token) { create(:oauth_access_token) }
  let(:context) { {viewer: user_1, doorkeeper_token: token, writable: true} }
  let(:reactable_id) { Canary::AnnictSchema.id_from_object(reactable, reactable.class) }
  let(:variables) { {reactableId: reactable_id, content: "HEART"} }
  let(:query) do
    <<~GRAPHQL
      mutation($reactableId: ID!, $content: ReactionContent!) {
        addReaction(input:{ reactableId: $reactableId, content: $content }) {
          reaction {
            user {
              username
            }
            content
            createdAt
          }
        }
      }
    GRAPHQL
  end
  let(:email_notification) { class_double("Deprecated::EmailNotificationService").as_stubbed_const }

  context "記録 (エピソード) にリアクションしたとき" do
    let(:episode_record) { create(:episode_record, user: user_2) }
    let(:reactable) { episode_record.record }

    it "Likeが保存されること" do
      expect(Like.count).to eq 0
      expect(episode_record.likes_count).to eq 0
      expect(Notification.count).to eq 0
      expect(email_notification).to receive(:send_email).with("liked_episode_record", user_1, user_1.id, episode_record.id)

      Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Like.count).to eq 1
      expect(episode_record.reload.likes_count).to eq 1
      expect(Notification.count).to eq 1

      like = Like.first

      expect(like.recipient).to eq episode_record
      expect(like.user).to eq user_1
    end
  end

  context "記録 (アニメ) にリアクションしたとき" do
    let(:work_record) { create(:work_record, user: user_2) }
    let(:reactable) { work_record.record }

    it "Likeが保存されること" do
      expect(Like.count).to eq 0
      expect(work_record.likes_count).to eq 0
      expect(Notification.count).to eq 0
      # アニメへの記録についたリアクションの通知は現状サポートしていない (サポートしたい)
      email_notification.should_receive(:send_email).exactly(0).times

      Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Like.count).to eq 1
      expect(work_record.reload.likes_count).to eq 1
      expect(Notification.count).to eq 1

      like = Like.first

      expect(like.recipient).to eq work_record
      expect(like.user).to eq user_1
    end
  end

  context "ステータス変更にリアクションしたとき" do
    let(:status) { create(:status, user: user_2) }
    let(:reactable) { status }

    it "Likeが保存されること" do
      expect(Like.count).to eq 0
      expect(status.likes_count).to eq 0
      expect(Notification.count).to eq 0
      # ステータスについたリアクションの通知は現状サポートしていない (サポートしたい)
      email_notification.should_receive(:send_email).exactly(0).times

      Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Like.count).to eq 1
      expect(status.reload.likes_count).to eq 1
      expect(Notification.count).to eq 1

      like = Like.first

      expect(like.recipient).to eq status
      expect(like.user).to eq user_1
    end
  end

  context "自分のリソースにリアクションしたとき" do
    let(:episode_record) { create(:episode_record, user: user_1) } # GraphQL APIの viewer と同じユーザ (= 自分)
    let(:reactable) { episode_record.record }

    it "Likeは保存されるが通知はされないこと" do
      expect(Like.count).to eq 0
      expect(episode_record.likes_count).to eq 0
      expect(Notification.count).to eq 0
      # メールは送らない
      email_notification.should_receive(:send_email).exactly(0).times

      Canary::AnnictSchema.execute(query, variables: variables, context: context)

      expect(Like.count).to eq 1
      expect(episode_record.reload.likes_count).to eq 1
      # 通知ページに表示しない
      expect(Notification.count).to eq 0

      like = Like.first

      expect(like.recipient).to eq episode_record
      expect(like.user).to eq user_1
    end
  end
end
