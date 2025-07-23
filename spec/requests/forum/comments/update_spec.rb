# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /forum/posts/:post_id/comments/:comment_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post)

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: "更新されたコメント"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(forum_comment.reload.body).not_to eq("更新されたコメント")
  end

  it "ログインしているとき、自分のコメントを正常に更新できること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user, body: "元のコメント")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: "更新されたコメント内容です。\n\n編集しました。"
      }
    }

    expect(forum_comment.reload.body).to eq("更新されたコメント内容です。\n\n編集しました。")
    expect(forum_comment.locale).to eq("ja")
    expect(response).to redirect_to(forum_post_path(forum_post))
    expect(flash[:notice]).to eq("更新しました")
  end

  it "ログインしているとき、英語の本文に更新すると、ロケールがenになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user, body: "日本語のコメント", locale: "ja")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: "This comment has been updated to English."
      }
    }

    expect(forum_comment.reload.body).to eq("This comment has been updated to English.")
    expect(forum_comment.locale).to eq("en")
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、他のユーザーのコメントを更新しようとすると、403エラーになること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: other_user)
    login_as(user, scope: :user)

    expect {
      patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
        forum_comment: {
          body: "他人のコメントを編集"
        }
      }
    }.to raise_error(Pundit::NotAuthorizedError)

    expect(forum_comment.reload.body).not_to eq("他人のコメントを編集")
  end

  it "ログインしているとき、本文が空の場合、エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user, body: "元のコメント")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: ""
      }
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("を入力してください")
    expect(forum_comment.reload.body).to eq("元のコメント")
  end

  it "ログインしているとき、本文が5000文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user, body: "元のコメント")
    login_as(user, scope: :user)

    long_body = "あ" * 5001

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: long_body
      }
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("5000文字以内で入力してください")
    expect(forum_comment.reload.body).to eq("元のコメント")
  end

  it "ログインしているとき、存在しない投稿IDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    forum_comment = create(:forum_comment, user: user)
    login_as(user, scope: :user)

    expect {
      patch "/forum/posts/99999/comments/#{forum_comment.id}", params: {
        forum_comment: {
          body: "存在しない投稿の更新"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、存在しないコメントIDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    login_as(user, scope: :user)

    expect {
      patch "/forum/posts/#{forum_post.id}/comments/99999", params: {
        forum_comment: {
          body: "存在しないコメントの更新"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "指定された投稿に属さないコメントIDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    forum_post1 = create(:forum_post)
    forum_post2 = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post1, user: user)
    login_as(user, scope: :user)

    expect {
      patch "/forum/posts/#{forum_post2.id}/comments/#{forum_comment.id}", params: {
        forum_comment: {
          body: "別の投稿のコメントを更新"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "コメント更新時にlast_commented_atは更新されないこと" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user)
    original_last_commented_at = forum_post.last_commented_at
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}", params: {
      forum_comment: {
        body: "更新されたコメント"
      }
    }

    expect(forum_post.reload.last_commented_at).to be_within(1.second).of(original_last_commented_at)
  end
end
