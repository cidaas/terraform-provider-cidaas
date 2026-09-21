# In the import statement, the identifier is the combination of consent_id,
# consent_version_id and locale joined by the special character ":"
# To import a consent version for multiple locales, you need to append the locales separated by ":".
# For example, "3f453233-92d4-475b-b10e:813fbd47-6c50-4fc4-881a:en-us:de:en" imports
# the consent version for the locales en-us, de and en.

terraform import cidaas_consent_version.v1 3f453233-92d4-475b-b10e:813fbd47-6c50-4fc4-881a:en-us