namespace :sync do
  task s3: :environment do
    AWS.config(
      access_key_id:     ENV['S3_ACCESS_KEY_ID'],
      secret_access_key: ENV['S3_SECRET_ACCESS_KEY']
    )

    development_bucket = ENV['S3_BUCKET_NAME_DEVELOPMENT']
    production_bucket  = ENV['S3_BUCKET_NAME_PRODUCTION']

    s3 = AWS::S3.new

    s3.buckets[development_bucket].clear!

    s3.buckets[production_bucket].objects.each do |obj|
      puts obj.key
      obj.copy_to(obj.key, bucket_name: development_bucket)
    end
  end
end