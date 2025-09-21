# typed: false
# frozen_string_literal: true

RSpec.describe "POST /forum/posts/:post_id/comments", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    forum_post = create(:forum_post, user: create(:registered_user))

    post "/forum/posts/#{forum_post.id}/comments", params: {
      forum_comment: {
        body: "コメント本文"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(ForumComment.count).to eq(0)
  end

  it "ログインしているとき、正常なパラメータでコメントが作成されること" do
    user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))
    login_as(user, scope: :user)

    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: "これはテストコメントです。\n\n投稿に対する返信です。"
        }
      }
    }.to change(ForumComment, :count).by(1)
      .and change(ForumPostParticipant, :count).by(1)

    comment = ForumComment.last
    expect(comment.user).to eq(user)
    expect(comment.forum_post).to eq(forum_post)
    expect(comment.body).to eq("これはテストコメントです。\n\n投稿に対する返信です。")
    expect(comment.locale).to eq("ja")

    participant = ForumPostParticipant.last
    expect(participant.forum_post).to eq(forum_post)
    expect(participant.user).to eq(user)

    expect(forum_post.reload.last_commented_at).to be_present

    expect(response).to redirect_to(forum_post_path(forum_post))
    expect(flash[:notice]).to eq("投稿しました")
  end

  it "ログインしているとき、英語の本文でコメントを作成すると、ロケールがenになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))
    login_as(user, scope: :user)

    post "/forum/posts/#{forum_post.id}/comments", params: {
      forum_comment: {
        body: "This is a test comment written in English."
      }
    }

    comment = ForumComment.last
    expect(comment.locale).to eq("en")
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、本文が空の場合、エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))
    login_as(user, scope: :user)

    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: ""
        }
      }
    }.not_to change(ForumComment, :count)

    expect(response.status).to eq(200)
    expect(response.body).to include("を入力してください")
  end

  it "ログインしているとき、本文が5000文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))
    login_as(user, scope: :user)

    long_body = "あ" * 5001

    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: long_body
        }
      }
    }.not_to change(ForumComment, :count)

    expect(response.status).to eq(200)
    expect(response.body).to include("5000文字以内で入力してください")
  end

  it "ログインしているとき、存在しない投稿IDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    post "/forum/posts/99999/comments", params: {
    forum_comment: {
    body: "存在しない投稿へのコメント"

    expect(response.status).to eq(404)
  end

  it "既に参加者である場合、重複した参加者レコードが作成されないこと" do
    user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))
    login_as(user, scope: :user)

    # 最初のコメント投稿で参加者になる
    post "/forum/posts/#{forum_post.id}/comments", params: {
      forum_comment: {
        body: "最初のコメント"
      }
    }

    expect(ForumPostParticipant.where(forum_post:, user:).count).to eq(1)

    # 2回目のコメント投稿
    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: "2回目のコメント"
        }
      }
    }.to change(ForumComment, :count).by(1)
      .and change(ForumPostParticipant, :count).by(0)

    expect(ForumPostParticipant.where(forum_post:, user:).count).to eq(1)
  end

  it "コメント作成時に通知が送信されること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    forum_post = create(:forum_post, user: other_user)

    # other_userを参加者として登録
    create(:forum_post_participant, forum_post:, user: other_user)

    login_as(user, scope: :user)

    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: "通知をテストするコメント"
        }
      }
    }.to have_enqueued_mail(ForumMailer, :comment_notification)
  end

  it "自分が既に参加している投稿にコメントした場合、自分には通知が送信されないこと" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    forum_post = create(:forum_post, user: create(:registered_user))

    # 両ユーザーを参加者として登録
    create(:forum_post_participant, forum_post:, user:)
    create(:forum_post_participant, forum_post:, user: other_user)

    login_as(user, scope: :user)

    # 通知をチェック - 他のユーザーには送信され、自分には送信されない
    expect {
      post "/forum/posts/#{forum_post.id}/comments", params: {
        forum_comment: {
          body: "自分には通知されないコメント"
        }
      }
    }.to have_enqueued_mail(ForumMailer, :comment_notification).exactly(:once)

    # 通知が他のユーザーにのみ送信されたことを確認
    enqueued_jobs = ActiveJob::Base.queue_adapter.enqueued_jobs
    mail_jobs = enqueued_jobs.select { |job|
      job[:job] == ActionMailer::MailDeliveryJob &&
        job[:args][0] == "ForumMailer" &&
        job[:args][1] == "comment_notification"
    }

    # 通知は1件のみ（other_userへの通知）
    expect(mail_jobs.size).to eq(1)
    expect(mail_jobs.first[:args][3]["args"][0]).to eq(other_user.id)
  end
end
